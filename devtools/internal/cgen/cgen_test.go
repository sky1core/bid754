package cgen

import (
	"bytes"
	"math/big"
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

func TestRenderBidgoUsesNamedUint128LimbsOnly(t *testing.T) {
	tables := []Table{
		{
			Spec:  TableSpec{Name: "bid_u128"},
			CType: "BID_UINT128",
			Dims:  []int{1},
			Value: Value{Elements: []Value{{Elements: []Value{{Number: big.NewInt(1)}, {Number: big.NewInt(2)}}}}},
		},
		{
			Spec:  TableSpec{Name: "bid_u256"},
			CType: "BID_UINT256",
			Dims:  []int{1},
			Value: Value{Elements: []Value{{Elements: []Value{
				{Number: big.NewInt(3)}, {Number: big.NewInt(4)}, {Number: big.NewInt(5)}, {Number: big.NewInt(6)},
			}}}},
		},
		{
			Spec:  TableSpec{Name: "bid_exp"},
			CType: "int",
			Dims:  []int{1},
			Value: Value{Elements: []Value{{Number: big.NewInt(-7)}}},
		},
	}

	generated, err := renderBidgo(tables)
	if err != nil {
		t.Fatalf("renderBidgo error: %v", err)
	}
	source := string(generated)
	for _, want := range []string{
		"var bid_u128 = [1]BID_UINT128{\n\t{lo: 1, hi: 2},\n}",
		"var bid_u256 = [1]BID_UINT256{\n\t{w: [4]uint64{3, 4, 5, 6}},\n}",
		"var bid_exp = [1]int{\n\t-7,\n}",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated bidgo tables missing %q:\n%s", want, source)
		}
	}
	if strings.Contains(source, "BID_UINT128{w:") || strings.Contains(source, "[2]uint64") {
		t.Fatalf("generated bidgo BID_UINT128 retained array representation:\n%s", source)
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
	bidgoPath := filepath.Join(repoRoot, manifest.BidgoOutput)

	assertFileMatches(t, goPath, generated.Go)
	assertFileMatches(t, rustPath, generated.Rust)
	assertFileMatches(t, bidgoPath, generated.Bidgo)

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

func TestEnsureUnconditionalDeclarationAcceptsPlainDecl(t *testing.T) {
	src := "static const int foo[3] = {1, 2, 3};\n"
	if err := ensureUnconditionalDeclaration(src, "foo"); err != nil {
		t.Fatalf("unexpected rejection: %v", err)
	}
}

func TestEnsureUnconditionalDeclarationIgnoresClosedConditional(t *testing.T) {
	// A conditional that opens and closes before the declaration (the shape of
	// the real Intel sources, where #if BID_BIG_ENDIAN checks an unrelated
	// typedef ahead of the tables) must not be flagged.
	src := "#if BID_BIG_ENDIAN\ntypedef int x;\n#else\ntypedef int y;\n#endif\nstatic const int foo[3] = {1, 2, 3};\n"
	if err := ensureUnconditionalDeclaration(src, "foo"); err != nil {
		t.Fatalf("unexpected rejection: %v", err)
	}
}

func TestEnsureUnconditionalDeclarationRejectsCheckedDecl(t *testing.T) {
	// Table defined only inside an #if branch: c-tablegen cannot tell which
	// endian/config branch it is and must reject unresolved input.
	src := "#if BID_BIG_ENDIAN\nstatic const int foo[3] = {3, 2, 1};\n#else\n#endif\n"
	if err := ensureUnconditionalDeclaration(src, "foo"); err == nil {
		t.Fatal("expected conditional-check rejection")
	}
}

func TestEnsureUnconditionalDeclarationRejectsEndianDoubleDefinition(t *testing.T) {
	src := "#if BID_BIG_ENDIAN\nstatic const int foo[3] = {3, 2, 1};\n#else\nstatic const int foo[3] = {1, 2, 3};\n#endif\n"
	if err := ensureUnconditionalDeclaration(src, "foo"); err == nil {
		t.Fatal("expected duplicate-declaration rejection")
	}
}

func TestPreprocessorConditionalDepthNesting(t *testing.T) {
	src := "#if A\n#if B\nX\n#endif\nY\n#endif\nZ\n"
	for _, tc := range []struct {
		marker string
		want   int
	}{
		{"X", 2}, {"Y", 1}, {"Z", 0},
	} {
		off := strings.Index(src, tc.marker)
		if got := preprocessorConditionalDepthAt(src, off); got != tc.want {
			t.Fatalf("depth at %q = %d, want %d", tc.marker, got, tc.want)
		}
	}
}

func TestParseTableFileEnforcesExpectedShape(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "table.c"), []byte("static const int foo[3] = {1, 2, 3};\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	base := TableSpec{Name: "foo", Source: "table.c", GoName: "Foo", RustName: "FOO"}

	good := base
	good.ExpectedShape = []int{3}
	table, err := ParseTableFile(dir, good)
	if err != nil {
		t.Fatalf("correct shape rejected: %v", err)
	}
	if !equalInts(table.Dims, []int{3}) {
		t.Fatalf("Dims = %v, want [3]", table.Dims)
	}

	bad := base
	bad.ExpectedShape = []int{4}
	if _, err := ParseTableFile(dir, bad); err == nil {
		t.Fatal("expected shape-mismatch rejection")
	}
}

func TestLoadManifestRequiresPositiveExpectedShape(t *testing.T) {
	dir := t.TempDir()
	for _, tc := range []struct {
		name  string
		table string
	}{
		{name: "missing", table: `{"name":"foo","source":"s","go_name":"Foo","rust_name":"FOO"}`},
		{name: "empty", table: `{"name":"foo","source":"s","go_name":"Foo","rust_name":"FOO","expected_shape":[]}`},
		{name: "zero", table: `{"name":"foo","source":"s","go_name":"Foo","rust_name":"FOO","expected_shape":[0]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, tc.name+".json")
			body := `{"go_package":"p","go_output":"g","rust_output":"r","bidgo_output":"b","bidgo_source":"s","tables":[` + tc.table + `]}`
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				t.Fatalf("write manifest: %v", err)
			}
			if _, err := LoadManifest(path); err == nil {
				t.Fatal("expected expected_shape validation error")
			}
		})
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
