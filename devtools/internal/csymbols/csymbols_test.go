package csymbols

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEvalExpr(t *testing.T) {
	macros := map[string]string{
		"DECIMAL_CALL_BY_REFERENCE":            "0",
		"DECIMAL_GLOBAL_ROUNDING":              "0",
		"DECIMAL_ALTERNATE_EXCEPTION_HANDLING": "0",
	}
	if !evalExpr("!DECIMAL_GLOBAL_ROUNDING && !DECIMAL_CALL_BY_REFERENCE", macros) {
		t.Fatal("expected expression to evaluate true")
	}
	if evalExpr("DECIMAL_CALL_BY_REFERENCE", macros) {
		t.Fatal("expected expression to evaluate false")
	}
}

func TestSplitParams(t *testing.T) {
	params := splitParams("BID_UINT32 x, BID_UINT32 y, _IDEC_round rnd_mode")
	want := []string{"BID_UINT32 x", "BID_UINT32 y", "_IDEC_round rnd_mode"}
	if len(params) != len(want) {
		t.Fatalf("len(params) = %d, want %d", len(params), len(want))
	}
	for i := range want {
		if params[i] != want[i] {
			t.Fatalf("params[%d] = %q, want %q", i, params[i], want[i])
		}
	}
}

func TestGeneratedSymbolsStayInSync(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	manifest, err := LoadManifest(filepath.Join(repoRoot, "symbolgen_manifest.json"))
	if err != nil {
		t.Fatalf("LoadManifest error: %v", err)
	}
	generated, err := Generate(repoRoot, manifest)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	path := filepath.Join(repoRoot, manifest.Output)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error: %v", path, err)
	}
	if !bytes.Equal(got, generated.JSON) {
		t.Fatalf("%s is out of date; run `go run ./cmd/c-symbolgen -manifest symbolgen_manifest.json`", path)
	}
}

func TestExtractedSymbolsIncludeKnownBindings(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	manifest, err := LoadManifest(filepath.Join(repoRoot, "symbolgen_manifest.json"))
	if err != nil {
		t.Fatalf("LoadManifest error: %v", err)
	}
	generated, err := Generate(repoRoot, manifest)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	if !bytes.Contains(generated.JSON, []byte(`"name": "bid32_add"`)) {
		t.Fatal("missing bid32_add")
	}
	if !bytes.Contains(generated.JSON, []byte(`"link_name": "__bid32_add"`)) {
		t.Fatal("missing __bid32_add alias")
	}
	if !bytes.Contains(generated.JSON, []byte(`"_IDEC_round rnd_mode"`)) {
		t.Fatal("missing expanded rounding parameter")
	}
	if !bytes.Contains(generated.JSON, []byte(`"name": "bid32_to_string"`)) {
		t.Fatal("missing bid32_to_string")
	}

	var payload SymbolFile
	if err := json.Unmarshal(generated.JSON, &payload); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	seen := map[string]bool{}
	for _, symbol := range payload.Symbols {
		key := symbol.Name + "\x00" + symbol.Declaration
		if seen[key] {
			t.Fatalf("duplicate symbol entry: %s", symbol.Name)
		}
		seen[key] = true
	}
}

func TestExtractedSymbolsCoverFutureBindingCandidates(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	manifest, err := LoadManifest(filepath.Join(repoRoot, "symbolgen_manifest.json"))
	if err != nil {
		t.Fatalf("LoadManifest error: %v", err)
	}
	generated, err := Generate(repoRoot, manifest)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	var payload SymbolFile
	if err := json.Unmarshal(generated.JSON, &payload); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	index := make(map[string]Symbol, len(payload.Symbols))
	for _, symbol := range payload.Symbols {
		index[symbol.Name] = symbol
	}

	cases := []struct {
		name       string
		returnType string
		parameters []string
	}{
		{name: "bid32_modf", returnType: "BID_UINT32", parameters: []string{"BID_UINT32 x", "BID_UINT32* y", "_IDEC_flags*pfpsf"}},
		{name: "bid64_modf", returnType: "BID_UINT64", parameters: []string{"BID_UINT64 x", "BID_UINT64* y", "_IDEC_flags*pfpsf"}},
		{name: "bid128_modf", returnType: "BID_UINT128", parameters: []string{"BID_UINT128 x", "BID_UINT128* y", "_IDEC_flags*pfpsf"}},
		{name: "bid32_nearbyint", returnType: "BID_UINT32", parameters: []string{"BID_UINT32 x", "_IDEC_round rnd_mode", "_IDEC_flags*pfpsf"}},
		{name: "bid64_nearbyint", returnType: "BID_UINT64", parameters: []string{"BID_UINT64 x", "_IDEC_round rnd_mode", "_IDEC_flags*pfpsf"}},
		{name: "bid128_nearbyint", returnType: "BID_UINT128", parameters: []string{"BID_UINT128 x", "_IDEC_round rnd_mode", "_IDEC_flags*pfpsf"}},
		{name: "bid32_nexttoward", returnType: "BID_UINT32", parameters: []string{"BID_UINT32 x", "BID_UINT128 y", "_IDEC_flags*pfpsf"}},
		{name: "bid64_nexttoward", returnType: "BID_UINT64", parameters: []string{"BID_UINT64 x", "BID_UINT128 y", "_IDEC_flags*pfpsf"}},
		{name: "bid128_nexttoward", returnType: "BID_UINT128", parameters: []string{"BID_UINT128 x", "BID_UINT128 y", "_IDEC_flags*pfpsf"}},
	}

	for _, tc := range cases {
		symbol, ok := index[tc.name]
		if !ok {
			t.Fatalf("missing symbol %q in extracted inventory", tc.name)
		}
		if symbol.ReturnType != tc.returnType {
			t.Fatalf("%s return type = %q, want %q", tc.name, symbol.ReturnType, tc.returnType)
		}
		if len(symbol.Parameters) != len(tc.parameters) {
			t.Fatalf("%s parameter count = %d, want %d", tc.name, len(symbol.Parameters), len(tc.parameters))
		}
		for i, want := range tc.parameters {
			if symbol.Parameters[i] != want {
				t.Fatalf("%s parameter[%d] = %q, want %q", tc.name, i, symbol.Parameters[i], want)
			}
		}
	}
}
