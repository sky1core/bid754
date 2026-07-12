package gotypegen

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratedTypesStayInSync(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	manifest, err := LoadManifest(filepath.Join(repoRoot, "typegen_manifest.json"))
	if err != nil {
		t.Fatalf("LoadManifest error: %v", err)
	}
	generated, err := Generate(manifest)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	path := filepath.Join(repoRoot, manifest.Output)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error: %v", path, err)
	}
	if !bytes.Equal(got, generated) {
		t.Fatalf("%s is out of date; run `go run ./cmd/go-typegen -manifest typegen_manifest.json`", path)
	}
}

func TestGeneratedTypesIncludeKnownDefinitions(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	manifest, err := LoadManifest(filepath.Join(repoRoot, "typegen_manifest.json"))
	if err != nil {
		t.Fatalf("LoadManifest error: %v", err)
	}
	generated, err := Generate(manifest)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	for _, snippet := range [][]byte{
		[]byte("type EncodingFormat int"),
		[]byte("EncodingBID EncodingFormat = iota"),
		[]byte("FlagInexact"),
		[]byte("ExceptionFlags = 1 << iota"),
		[]byte("type Decimal64BID uint64"),
		[]byte("type Decimal128BID [16]byte"),
		[]byte("func (d Decimal128BID) ToBytes() [16]byte"),
	} {
		if !bytes.Contains(generated, snippet) {
			t.Fatalf("generated definitions missing %q", snippet)
		}
	}
}
