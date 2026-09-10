package testgen

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestFinitePathsOutputsReproduce(t *testing.T) {
	files, err := GenerateFinitePathsOutputs()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("finite path artifacts=%d, want Go and Rust", len(files))
	}
	for path, generated := range files {
		actual, err := os.ReadFile(filepath.Join("..", "..", path))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(actual, generated) {
			t.Errorf("stale generated finite executor: %s", path)
		}
	}
}
