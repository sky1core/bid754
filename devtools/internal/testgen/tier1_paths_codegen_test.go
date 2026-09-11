package testgen

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestTier1PathsOutputsReproduce(t *testing.T) {
	outputs, err := GenerateTier1Outputs()
	if err != nil {
		t.Fatal(err)
	}
	if len(outputs) != 3 {
		t.Fatalf("artifacts=%d, want Go, Rust and Java", len(outputs))
	}
	for _, path := range []string{tier1PathsGoPath, tier1PathsRustPath, "../devtools/java/Tier1BigDecimalProbe.java"} {
		generated, ok := outputs[path]
		if !ok {
			t.Fatalf("missing %s", path)
		}
		fullPath := filepath.Join("..", "..", path)

		actual, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(actual, generated) {
			t.Errorf("stale generated Tier1 executor: %s", path)
		}
	}
}
