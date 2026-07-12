package testgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPublicAPISigsRejectsBuildVariantKindConflict(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"a_mutable.go":  "//go:build mutable\n\npackage bid754\n\nvar Zero64BID Decimal64BID\n",
		"z_accessor.go": "//go:build !mutable\n\npackage bid754\n\nfunc Zero64BID() Decimal64BID { return 0 }\n",
	}
	for name, source := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	_, err := loadPublicAPISigs(dir)
	if err == nil || !strings.Contains(err.Error(), "conflicting declarations") {
		t.Fatalf("loadPublicAPISigs(build-tag var/func conflict) error = %v, want conflict rejection", err)
	}
}

func TestLoadPublicAPISigsAllowsExactBuildVariantDuplicate(t *testing.T) {
	dir := t.TempDir()
	source := "package bid754\n\nfunc Transform(values map[string]int) map[string]int { return values }\n"
	for _, name := range []string{"a.go", "z.go"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	symbols, err := loadPublicAPISigs(dir)
	if err != nil {
		t.Fatalf("loadPublicAPISigs(exact duplicate): %v", err)
	}
	if len(symbols) != 1 || symbols[0].Symbol != "Transform" {
		t.Fatalf("exact duplicate symbols = %+v, want one Transform entry", symbols)
	}
}

func TestLoadPublicAPISigsRejectsBuildVariantSignatureConflicts(t *testing.T) {
	tests := []struct {
		name string
		a    string
		z    string
	}{
		{
			name: "map types",
			a:    "package bid754\n\nfunc Transform(values map[string]int) {}\n",
			z:    "package bid754\n\nfunc Transform(values map[int]string) {}\n",
		},
		{
			name: "receiver pointer",
			a:    "package bid754\n\ntype Decimal64BID uint64\n\nfunc (d Decimal64BID) Transform() {}\n",
			z:    "package bid754\n\ntype Decimal64BID uint64\n\nfunc (d *Decimal64BID) Transform() {}\n",
		},
		{
			name: "generic constraint",
			a:    "package bid754\n\nfunc Transform[T ~int](value T) {}\n",
			z:    "package bid754\n\nfunc Transform[T ~string](value T) {}\n",
		},
		{
			name: "generic parameter order",
			a:    "package bid754\n\nfunc Transform[T, U any](x T, y U) {}\n",
			z:    "package bid754\n\nfunc Transform[U, T any](x T, y U) {}\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, source := range map[string]string{"a.go": tc.a, "z.go": tc.z} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600); err != nil {
					t.Fatalf("write %s: %v", name, err)
				}
			}
			_, err := loadPublicAPISigs(dir)
			if err == nil || !strings.Contains(err.Error(), "conflicting declarations") {
				t.Fatalf("loadPublicAPISigs(signature conflict) error = %v, want conflict rejection", err)
			}
		})
	}
}

func TestResolvePublicAPIRejectsExportedMutableVariables(t *testing.T) {
	symbols := []publicAPISymbol{{Symbol: "Zero64BID", Kind: "var", Name: "Zero64BID"}}
	spec := &PublicAPISpec{SymbolExclusions: []PublicSymbolExclusion{{
		Symbol: "Zero64BID",
		Class:  "constant_accessor",
		Reason: "test exclusion must not make an exported mutable variable acceptable",
	}}}

	_, err := resolvePublicAPI(symbols, nil, spec)
	if err == nil || !strings.Contains(err.Error(), "exported mutable variable") {
		t.Fatalf("resolvePublicAPI(exported var) error = %v, want exported-mutable-variable rejection", err)
	}
}

func TestResolvePublicAPIAcceptsConstantAccessorFunction(t *testing.T) {
	symbols := []publicAPISymbol{{
		Symbol:  "Zero64BID",
		Kind:    "func",
		Name:    "Zero64BID",
		Results: []string{"Decimal64BID"},
	}}
	spec := &PublicAPISpec{SymbolExclusions: []PublicSymbolExclusion{{
		Symbol: "Zero64BID",
		Class:  "constant_accessor",
		Reason: "returns a copy of the package-private decimal constant",
	}}}

	inventory, err := resolvePublicAPI(symbols, nil, spec)
	if err != nil {
		t.Fatalf("resolvePublicAPI(constant accessor): %v", err)
	}
	if inventory.Total != 1 || inventory.Excluded != 1 || len(inventory.Symbols) != 1 {
		t.Fatalf("constant accessor inventory = %+v, want one excluded public function", inventory)
	}
	row := inventory.Symbols[0]
	if row.Kind != "func" || row.Status != "excluded_constant_accessor" {
		t.Fatalf("constant accessor row = %+v, want func/excluded_constant_accessor", row)
	}
}

func TestResolvePublicAPIRejectsGenericConstantAccessor(t *testing.T) {
	symbols := []publicAPISymbol{{
		Symbol:     "Zero64BID",
		Kind:       "func",
		Name:       "Zero64BID",
		TypeParams: []string{"T any"},
		Results:    []string{"Decimal64BID"},
	}}
	spec := &PublicAPISpec{SymbolExclusions: []PublicSymbolExclusion{{
		Symbol: "Zero64BID",
		Class:  "constant_accessor",
		Reason: "test generic accessors are not immutable constant plumbing",
	}}}

	_, err := resolvePublicAPI(symbols, nil, spec)
	if err == nil || !strings.Contains(err.Error(), "must be a non-generic package function") {
		t.Fatalf("resolvePublicAPI(generic constant accessor) error = %v, want exact accessor-shape rejection", err)
	}
}

func TestPublicParityFiniteCohortOracle(t *testing.T) {
	tests := []struct {
		input       string
		wantQuantum string
		wantDigits  int
		ok          bool
	}{
		{input: "1", wantQuantum: "0", wantDigits: 1, ok: true},
		{input: "1.25", wantQuantum: "-2", wantDigits: 3, ok: true},
		{input: ".25e3", wantQuantum: "1", wantDigits: 2, ok: true},
		{input: "1.e-4", wantQuantum: "-4", wantDigits: 1, ok: true},
		{input: " \t-1.0E+91", wantQuantum: "90", wantDigits: 2, ok: true},
		{input: "0001000.00", wantQuantum: "-2", wantDigits: 6, ok: true},
		{input: "0.000", wantQuantum: "-3", wantDigits: 0, ok: true},
		{input: "1000000.0", wantQuantum: "-1", wantDigits: 8, ok: true},
		{input: "1e9223372036854775808", wantQuantum: "9223372036854775808", wantDigits: 1, ok: true},
		{input: "1e-9223372036854775809", wantQuantum: "-9223372036854775809", wantDigits: 1, ok: true},
		{input: "Infinity"},
		{input: "1e"},
		{input: "1.0 "},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, ok := publicParityFiniteCohort(tc.input)
			if ok != tc.ok {
				t.Fatalf("publicParityFiniteCohort(%q) ok = %v, want %v", tc.input, ok, tc.ok)
			}
			if !ok {
				return
			}
			if got.Quantum.String() != tc.wantQuantum || got.CoefficientDigits != tc.wantDigits {
				t.Fatalf("publicParityFiniteCohort(%q) = (quantum=%s, digits=%d), want (%s, %d)", tc.input, got.Quantum, got.CoefficientDigits, tc.wantQuantum, tc.wantDigits)
			}
		})
	}
}

func TestPublicParityStringCorpusMetadataMatchesIndependentOracles(t *testing.T) {
	if err := validatePublicParityStringCorpus(); err != nil {
		t.Fatal(err)
	}
}
