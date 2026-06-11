package cgen

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseInitializerEvaluatesExpressions(t *testing.T) {
	value, err := parseInitializer("{1, 65 - 64, {0x10ull, -2}}")
	if err != nil {
		t.Fatalf("parseInitializer error: %v", err)
	}
	if got := value.Elements[0].Number.String(); got != "1" {
		t.Fatalf("value[0] = %s, want 1", got)
	}
	if got := value.Elements[1].Number.String(); got != "1" {
		t.Fatalf("value[1] = %s, want 1", got)
	}
	if got := value.Elements[2].Elements[0].Number.String(); got != "16" {
		t.Fatalf("value[2][0] = %s, want 16", got)
	}
	if got := value.Elements[2].Elements[1].Number.String(); got != "-2" {
		t.Fatalf("value[2][1] = %s, want -2", got)
	}

	shape, err := inferShape(value)
	if err == nil {
		t.Fatalf("expected ragged initializer error, got shape %v", shape)
	}
}

func TestParseTableFileBidShortRecipScale(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	table, err := ParseTableFile(repoRoot, TableSpec{
		Name:     "bid_short_recip_scale",
		Source:   "third_party/intel_dfp/src/bid_decimal_data.c",
		GoName:   "BidShortRecipScale",
		RustName: "BID_SHORT_RECIP_SCALE",
	})
	if err != nil {
		t.Fatalf("ParseTableFile error: %v", err)
	}

	if got, want := table.CType, "int"; got != want {
		t.Fatalf("CType = %q, want %q", got, want)
	}
	if got, want := table.Dims, []int{18}; !equalInts(got, want) {
		t.Fatalf("Dims = %v, want %v", got, want)
	}
	if got := table.Value.Elements[1].Number.String(); got != "1" {
		t.Fatalf("table.Value[1] = %s, want 1", got)
	}
	if got := table.Value.Elements[17].Number.String(); got != "54" {
		t.Fatalf("table.Value[17] = %s, want 54", got)
	}
}

func TestGoScalarForCIntUsesFixedWidth(t *testing.T) {
	if got, want := goScalarFor("int"), "int32"; got != want {
		t.Fatalf("goScalarFor(%q) = %q, want %q", "int", got, want)
	}
}

func TestParseTableFileBidNrDigits(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	table, err := ParseTableFile(repoRoot, TableSpec{
		Name:     "bid_nr_digits",
		Source:   "third_party/intel_dfp/src/bid128.c",
		GoName:   "BidNrDigits",
		RustName: "BID_NR_DIGITS",
	})
	if err != nil {
		t.Fatalf("ParseTableFile error: %v", err)
	}

	if got, want := table.CType, "DEC_DIGITS"; got != want {
		t.Fatalf("CType = %q, want %q", got, want)
	}
	if len(table.Dims) != 1 {
		t.Fatalf("Dims len = %d, want 1", len(table.Dims))
	}
	first := table.Value.Elements[0]
	if got := first.Elements[0].Number.String(); got != "1" {
		t.Fatalf("digits = %s, want 1", got)
	}
	if got := first.Elements[2].Number.String(); got != "10" {
		t.Fatalf("threshold_lo = %s, want 10", got)
	}
}

func TestParseTableFileBidMidiTbl(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	table, err := ParseTableFile(repoRoot, TableSpec{
		Name:     "bid_midi_tbl",
		Source:   "third_party/intel_dfp/src/bid128_2_str_tables.c",
		GoName:   "BidMidiTbl",
		RustName: "BID_MIDI_TBL",
	})
	if err != nil {
		t.Fatalf("ParseTableFile error: %v", err)
	}

	if got, want := table.CType, "char"; got != want {
		t.Fatalf("CType = %q, want %q", got, want)
	}
	if got, want := table.Dims, []int{1000, 3}; !equalInts(got, want) {
		t.Fatalf("Dims = %v, want %v", got, want)
	}
	first := table.Value.Elements[0]
	if got := string([]byte{
		byte(first.Elements[0].Number.Int64()),
		byte(first.Elements[1].Number.Int64()),
		byte(first.Elements[2].Number.Int64()),
	}); got != "000" {
		t.Fatalf("first entry = %q, want %q", got, "000")
	}
}

func TestParseTableFileBidEx64m64(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	table, err := ParseTableFile(repoRoot, TableSpec{
		Name:     "bid_Ex64m64",
		Source:   "third_party/intel_dfp/src/bid128.c",
		GoName:   "BidEx64M64",
		RustName: "BID_EX64M64",
	})
	if err != nil {
		t.Fatalf("ParseTableFile error: %v", err)
	}

	if got, want := table.CType, "unsigned int"; got != want {
		t.Fatalf("CType = %q, want %q", got, want)
	}
	if got, want := table.Dims, []int{17}; !equalInts(got, want) {
		t.Fatalf("Dims = %v, want %v", got, want)
	}
	if got := table.Value.Elements[0].Number.String(); got != "3" {
		t.Fatalf("value[0] = %s, want 3", got)
	}
}

func TestParseTableFileBidMidpoint192(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	table, err := ParseTableFile(repoRoot, TableSpec{
		Name:     "bid_midpoint192",
		Source:   "third_party/intel_dfp/src/bid128.c",
		GoName:   "BidMidpoint192",
		RustName: "BID_MIDPOINT192",
	})
	if err != nil {
		t.Fatalf("ParseTableFile error: %v", err)
	}

	if got, want := table.CType, "BID_UINT192"; got != want {
		t.Fatalf("CType = %q, want %q", got, want)
	}
	if got, want := table.Dims, []int{20}; !equalInts(got, want) {
		t.Fatalf("Dims = %v, want %v", got, want)
	}
	first := table.Value.Elements[0]
	if got := len(first.Elements); got != 3 {
		t.Fatalf("first tuple len = %d, want 3", got)
	}
	if got := first.Elements[2].Number.String(); got != "1" {
		t.Fatalf("high word = %s, want 1", got)
	}
}

func TestParseTableFileBidTen2K256(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	table, err := ParseTableFile(repoRoot, TableSpec{
		Name:     "bid_ten2k256",
		Source:   "third_party/intel_dfp/src/bid128.c",
		GoName:   "BidTen2K256",
		RustName: "BID_TEN2K256",
	})
	if err != nil {
		t.Fatalf("ParseTableFile error: %v", err)
	}

	if got, want := table.CType, "BID_UINT256"; got != want {
		t.Fatalf("CType = %q, want %q", got, want)
	}
	if got, want := table.Dims, []int{39}; !equalInts(got, want) {
		t.Fatalf("Dims = %v, want %v", got, want)
	}
	first := table.Value.Elements[0]
	if got := len(first.Elements); got != 4 {
		t.Fatalf("first tuple len = %d, want 4", got)
	}
	if got := first.Elements[2].Number.String(); got != "2" {
		t.Fatalf("word[2] = %s, want 2", got)
	}
}

func TestGeneratedArtifactsStayInSync(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	manifest, err := LoadManifest(filepath.Join(repoRoot, "tablegen_manifest.json"))
	if err != nil {
		t.Fatalf("LoadManifest error: %v", err)
	}
	generated, err := Generate(repoRoot, manifest)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	goPath := filepath.Join(repoRoot, manifest.GoOutput)
	rustPath := filepath.Join(repoRoot, manifest.RustOutput)

	assertFileMatches(t, goPath, generated.Go)
	assertFileMatches(t, rustPath, generated.Rust)

	goSource := string(generated.Go)
	for _, want := range []string{
		"var BidEstimateDecimalDigits = [129]int32{",
		"var BidShortRecipScale = [18]int32{",
	} {
		if !strings.Contains(goSource, want) {
			t.Fatalf("generated Go tables missing %q", want)
		}
	}
}

func assertFileMatches(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error: %v", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s is out of date; run `go run ./cmd/c-tablegen -manifest tablegen_manifest.json`", path)
	}
}
