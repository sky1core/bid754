package testgen

// Hand-written cross-language corpus anchor for the Tier 1 long runners.
//
// The boundary tables are injected into both language runners from one
// generator source, but the probe sets, the ScaleB structured exponent list,
// and the quiet-comparison semantic matrix are emitted as per-language
// template literals. A drift in one template keeps both runners green against
// Intel C (each side still runs a self-consistent corpus), silently breaking
// the "same deterministic corpus for Go and Rust" contract
// (docs/TEST_GENERATION_SPEC.md). This test re-derives those literals from
// the checked-in generated artifacts and fails on any divergence. It is a
// hand-maintained anchor outside the generation path; generators must not
// write it (GUARDRAILS.md).

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const (
	tier1CrosscheckProbeCount        = 12
	tier1CrosscheckScaleExponents    = 25
	tier1CrosscheckQuietMatrixRows   = 16
	tier1CrosscheckQuietMatrixWidths = 3
)

func tier1CrosscheckRead(t *testing.T, parts ...string) string {
	t.Helper()
	path := filepath.Join(append([]string{"..", "..", ".."}, parts...)...)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

// tier1CrosscheckParseUint parses Go/Rust integer literal spellings that
// appear in the runner corpora, mapping the Rust i64::MIN/i64::MAX tokens to
// their two's-complement values so both languages normalize identically.
func tier1CrosscheckParseUint(t *testing.T, token string) uint64 {
	t.Helper()
	token = strings.TrimSpace(token)
	switch token {
	case "i64::MIN":
		return 0x8000000000000000
	case "i64::MAX":
		return 0x7fffffffffffffff
	}
	negative := strings.HasPrefix(token, "-")
	body := strings.TrimPrefix(token, "-")
	var value uint64
	var err error
	if strings.HasPrefix(body, "0x") || strings.HasPrefix(body, "0X") {
		value, err = strconv.ParseUint(body[2:], 16, 64)
	} else {
		value, err = strconv.ParseUint(body, 10, 64)
	}
	if err != nil {
		t.Fatalf("unparseable corpus literal %q: %v", token, err)
	}
	if negative {
		return -value
	}
	return value
}

var tier1CrosscheckTokenPattern = regexp.MustCompile(`i64::MIN|i64::MAX|-?0[xX][0-9a-fA-F]+|-?\d+`)

func tier1CrosscheckScalarList(t *testing.T, src, label string) []uint64 {
	t.Helper()
	tokens := tier1CrosscheckTokenPattern.FindAllString(src, -1)
	values := make([]uint64, 0, len(tokens))
	for _, token := range tokens {
		values = append(values, tier1CrosscheckParseUint(t, token))
	}
	if len(values) == 0 {
		t.Fatalf("%s: no corpus literals extracted", label)
	}
	return values
}

func tier1CrosscheckBlock(t *testing.T, src, label string, pattern *regexp.Regexp) string {
	t.Helper()
	match := pattern.FindStringSubmatch(src)
	if match == nil {
		t.Fatalf("%s: declaration not found (parser out of sync with the generated runner layout)", label)
	}
	return match[1]
}

// 128-bit corpus entries appear as {lo: X, hi: Y} in Go and
// Words { lo: X, hi: Y } in Rust; both reduce to lo/hi pairs.
var tier1CrosscheckWordsPattern = regexp.MustCompile(`\{\s*lo:\s*([^,\s]+),\s*hi:\s*([^,}\s]+)\s*\}`)

func tier1CrosscheckWordsList(t *testing.T, src, label string) [][2]uint64 {
	t.Helper()
	matches := tier1CrosscheckWordsPattern.FindAllStringSubmatch(src, -1)
	pairs := make([][2]uint64, 0, len(matches))
	for _, match := range matches {
		pairs = append(pairs, [2]uint64{
			tier1CrosscheckParseUint(t, match[1]),
			tier1CrosscheckParseUint(t, match[2]),
		})
	}
	if len(pairs) == 0 {
		t.Fatalf("%s: no lo/hi pairs extracted", label)
	}
	return pairs
}

func tier1CrosscheckEqualScalars(t *testing.T, label string, want, got []uint64) {
	t.Helper()
	if len(want) != len(got) {
		t.Fatalf("%s: length mismatch go=%d rust=%d", label, len(want), len(got))
	}
	for i := range want {
		if want[i] != got[i] {
			t.Fatalf("%s: divergence at index %d: go=%#x rust=%#x", label, i, want[i], got[i])
		}
	}
}

func tier1CrosscheckEqualWords(t *testing.T, label string, want, got [][2]uint64) {
	t.Helper()
	if len(want) != len(got) {
		t.Fatalf("%s: length mismatch go=%d rust=%d", label, len(want), len(got))
	}
	for i := range want {
		if want[i] != got[i] {
			t.Fatalf("%s: divergence at index %d: go=%#x rust=%#x", label, i, want[i], got[i])
		}
	}
}

func TestTier1RunnerCorporaAgreeAcrossLanguages(t *testing.T) {
	goArith := tier1CrosscheckRead(t, "bid754-go", "generated_ffi_bitcompare_tier1_arithmetic_long_test.go")
	goCompare := tier1CrosscheckRead(t, "bid754-go", "generated_ffi_bitcompare_tier1_compare_conversion_long_test.go")
	rustArith := tier1CrosscheckRead(t, "bid754-rs", "ffi-verify", "tests", "tier1_arithmetic_long_generated.rs")
	rustCompare := tier1CrosscheckRead(t, "bid754-rs", "ffi-verify", "tests", "tier1_compare_conversion_long_generated.rs")

	goProbes32 := tier1CrosscheckScalarList(t, tier1CrosscheckBlock(t, goArith, "go probes32",
		regexp.MustCompile(`(?s)var tier1ArithmeticProbes32 = \[\.\.\.\]uint32\{(.*?)\n\}`)), "go probes32")
	goProbes64 := tier1CrosscheckScalarList(t, tier1CrosscheckBlock(t, goArith, "go probes64",
		regexp.MustCompile(`(?s)var tier1ArithmeticProbes64 = \[\.\.\.\]uint64\{(.*?)\n\}`)), "go probes64")
	goProbes128 := tier1CrosscheckWordsList(t, tier1CrosscheckBlock(t, goArith, "go probes128",
		regexp.MustCompile(`(?s)var tier1ArithmeticProbes128 = \[\.\.\.\]tier1Arithmetic128Words\{(.*?)\n\}`)), "go probes128")

	if len(goProbes32) != tier1CrosscheckProbeCount || len(goProbes64) != tier1CrosscheckProbeCount || len(goProbes128) != tier1CrosscheckProbeCount {
		t.Fatalf("go probe counts %d/%d/%d, want %d each", len(goProbes32), len(goProbes64), len(goProbes128), tier1CrosscheckProbeCount)
	}

	// The Rust probe literals are duplicated per runner file; every copy must
	// match the Go corpus.
	for _, rust := range []struct {
		name string
		src  string
	}{
		{name: "rust arithmetic runner", src: rustArith},
		{name: "rust compare/conversion runner", src: rustCompare},
	} {
		probes32 := tier1CrosscheckScalarList(t, tier1CrosscheckBlock(t, rust.src, rust.name+" probes32",
			regexp.MustCompile(`(?s)const PROBES32: \[u32; \d+\] = \[(.*?)\];`)), rust.name+" probes32")
		probes64 := tier1CrosscheckScalarList(t, tier1CrosscheckBlock(t, rust.src, rust.name+" probes64",
			regexp.MustCompile(`(?s)const PROBES64: \[u64; \d+\] = \[(.*?)\];`)), rust.name+" probes64")
		probes128 := tier1CrosscheckWordsList(t, tier1CrosscheckBlock(t, rust.src, rust.name+" probes128",
			regexp.MustCompile(`(?s)const PROBES128: \[Words; \d+\] = \[(.*?)\];`)), rust.name+" probes128")
		tier1CrosscheckEqualScalars(t, rust.name+" probes32", goProbes32, probes32)
		tier1CrosscheckEqualScalars(t, rust.name+" probes64", goProbes64, probes64)
		tier1CrosscheckEqualWords(t, rust.name+" probes128", goProbes128, probes128)
	}

	goScale := tier1CrosscheckScalarList(t, tier1CrosscheckBlock(t, goArith, "go scale exponents",
		regexp.MustCompile(`(?s)var tier1ArithmeticScaleExponentValues = \[\.\.\.\]int64\{(.*?)\n\}`)), "go scale exponents")
	rustScale := tier1CrosscheckScalarList(t, tier1CrosscheckBlock(t, rustArith, "rust scale exponents",
		regexp.MustCompile(`(?s)const SCALE_EXPONENTS: \[i64; \d+\] = \[(.*?)\];`)), "rust scale exponents")
	if len(goScale) != tier1CrosscheckScaleExponents {
		t.Fatalf("go scale exponent count %d, want %d", len(goScale), tier1CrosscheckScaleExponents)
	}
	tier1CrosscheckEqualScalars(t, "scaleb structured exponents", goScale, rustScale)

	tier1CrosscheckQuietMatrix(t, goCompare, rustCompare)
}

type tier1CrosscheckQuietRow struct {
	xLo, xHi uint64
	yLo, yHi uint64
	relation int64
	invalid  bool
}

// Go rows: {name: "...", x: <lit>, y: <lit>, relation: N[, flags: FlagInvalidOperation]}
// with 128-bit operands spelled value(lo, hi).
var tier1CrosscheckGoQuietRow = regexp.MustCompile(
	`\{name: "[^"]*", x: (value\([^)]*\)|[0-9a-fx]+), y: (value\([^)]*\)|[0-9a-fx]+), relation: (-?\d+)(, flags: FlagInvalidOperation)?\}`)
var tier1CrosscheckGoValueCall = regexp.MustCompile(`value\(\s*([^,\s]+),\s*([^)\s]+)\s*\)`)

func tier1CrosscheckGoQuietOperand(t *testing.T, token string) (uint64, uint64) {
	t.Helper()
	if match := tier1CrosscheckGoValueCall.FindStringSubmatch(token); match != nil {
		return tier1CrosscheckParseUint(t, match[1]), tier1CrosscheckParseUint(t, match[2])
	}
	return tier1CrosscheckParseUint(t, token), 0
}

// Rust rows: (<operand>, <operand>, N, invalid|ExceptionFlags::empty()) where
// an operand is a literal, Words { lo: .., hi: .. }, or a local alias
// (let <name> = Words { .. };) declared above the matrix.
var tier1CrosscheckRustAlias = regexp.MustCompile(`let (\w+) = Words \{ lo: ([^,]+), hi: ([^}]+) \};`)
var tier1CrosscheckRustQuietRow = regexp.MustCompile(
	`\((Words \{[^}]*\}|[\w:]+), (Words \{[^}]*\}|[\w:]+), (-?\d+), (invalid|ExceptionFlags::empty\(\))\)`)

func tier1CrosscheckRustQuietOperand(t *testing.T, token string, aliases map[string][2]uint64) (uint64, uint64) {
	t.Helper()
	token = strings.TrimSpace(token)
	if pair, ok := aliases[token]; ok {
		return pair[0], pair[1]
	}
	if match := tier1CrosscheckWordsPattern.FindStringSubmatch(token); match != nil {
		return tier1CrosscheckParseUint(t, match[1]), tier1CrosscheckParseUint(t, match[2])
	}
	return tier1CrosscheckParseUint(t, token), 0
}

func tier1CrosscheckSection(t *testing.T, src, label, from, to string) string {
	t.Helper()
	start := strings.Index(src, from)
	if start < 0 {
		t.Fatalf("%s: start marker %q not found", label, from)
	}
	rest := src[start:]
	if to == "" {
		return rest
	}
	end := strings.Index(rest[len(from):], to)
	if end < 0 {
		t.Fatalf("%s: end marker %q not found", label, to)
	}
	return rest[:len(from)+end]
}

func tier1CrosscheckQuietMatrix(t *testing.T, goCompare, rustCompare string) {
	t.Helper()

	goMatrix := tier1CrosscheckSection(t, goCompare, "go quiet matrix",
		"func TestTier1QuietComparisonSemanticMatrix", "\nfunc ")
	goSections := []string{
		tier1CrosscheckSection(t, goMatrix, "go quiet matrix decimal32", `t.Run("decimal32"`, `t.Run("decimal64"`),
		tier1CrosscheckSection(t, goMatrix, "go quiet matrix decimal64", `t.Run("decimal64"`, `t.Run("decimal128"`),
		tier1CrosscheckSection(t, goMatrix, "go quiet matrix decimal128", `t.Run("decimal128"`, ""),
	}

	rustMatrix := tier1CrosscheckSection(t, rustCompare, "rust quiet matrix",
		"fn tier1_quiet_comparison_semantic_matrix", "\n#[test]")
	aliases := map[string][2]uint64{}
	for _, match := range tier1CrosscheckRustAlias.FindAllStringSubmatch(rustMatrix, -1) {
		aliases[match[1]] = [2]uint64{
			tier1CrosscheckParseUint(t, match[2]),
			tier1CrosscheckParseUint(t, match[3]),
		}
	}
	rustBlocks := []string{}
	for rest := rustMatrix; ; {
		start := strings.Index(rest, "for (x, y, relation, flags) in [")
		if start < 0 {
			break
		}
		rest = rest[start:]
		end := strings.Index(rest, "] {")
		if end < 0 {
			t.Fatal("rust quiet matrix: unterminated row block")
		}
		rustBlocks = append(rustBlocks, rest[:end])
		rest = rest[end:]
	}
	if len(rustBlocks) != tier1CrosscheckQuietMatrixWidths {
		t.Fatalf("rust quiet matrix: found %d row blocks, want %d", len(rustBlocks), tier1CrosscheckQuietMatrixWidths)
	}

	widths := [...]string{"decimal32", "decimal64", "decimal128"}
	for i, width := range widths {
		goRows := []tier1CrosscheckQuietRow{}
		for _, match := range tier1CrosscheckGoQuietRow.FindAllStringSubmatch(goSections[i], -1) {
			xLo, xHi := tier1CrosscheckGoQuietOperand(t, match[1])
			yLo, yHi := tier1CrosscheckGoQuietOperand(t, match[2])
			relation, err := strconv.ParseInt(match[3], 10, 64)
			if err != nil {
				t.Fatalf("go quiet matrix %s: bad relation %q", width, match[3])
			}
			goRows = append(goRows, tier1CrosscheckQuietRow{
				xLo: xLo, xHi: xHi, yLo: yLo, yHi: yHi,
				relation: relation, invalid: match[4] != "",
			})
		}
		rustRows := []tier1CrosscheckQuietRow{}
		for _, match := range tier1CrosscheckRustQuietRow.FindAllStringSubmatch(rustBlocks[i], -1) {
			xLo, xHi := tier1CrosscheckRustQuietOperand(t, match[1], aliases)
			yLo, yHi := tier1CrosscheckRustQuietOperand(t, match[2], aliases)
			relation, err := strconv.ParseInt(match[3], 10, 64)
			if err != nil {
				t.Fatalf("rust quiet matrix %s: bad relation %q", width, match[3])
			}
			rustRows = append(rustRows, tier1CrosscheckQuietRow{
				xLo: xLo, xHi: xHi, yLo: yLo, yHi: yHi,
				relation: relation, invalid: match[4] == "invalid",
			})
		}
		if len(goRows) != tier1CrosscheckQuietMatrixRows {
			t.Fatalf("go quiet matrix %s: %d rows, want %d", width, len(goRows), tier1CrosscheckQuietMatrixRows)
		}
		if len(rustRows) != tier1CrosscheckQuietMatrixRows {
			t.Fatalf("rust quiet matrix %s: %d rows, want %d", width, len(rustRows), tier1CrosscheckQuietMatrixRows)
		}
		for j := range goRows {
			if goRows[j] != rustRows[j] {
				t.Fatalf("quiet matrix %s row %d diverges: go=%+v rust=%+v", width, j, goRows[j], rustRows[j])
			}
		}
	}
}
