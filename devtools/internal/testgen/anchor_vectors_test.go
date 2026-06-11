package testgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestBidCodecVectorAnchorsExistInCheckedInVectors enforces the spec contract
// that every anchor record stays byte-for-byte present in the checked-in
// bid-codec-vectors/vectors.json. If a generator change drops or alters an
// anchor record, this test fails even though the consumer-side anchor checks
// (which are independent of vectors.json) would still pass.
func TestBidCodecVectorAnchorsExistInCheckedInVectors(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	raw, err := os.ReadFile(filepath.Join(repoRoot, "bid-codec-vectors", "vectors.json"))
	if err != nil {
		t.Fatalf("read checked-in vectors.json: %v", err)
	}
	var payload struct {
		FormatVersion int              `json:"format_version"`
		Vectors       []bidCodecVector `json:"vectors"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal vectors.json: %v", err)
	}
	if payload.FormatVersion != bidCodecVectorFormatVersion {
		t.Fatalf("vectors.json format_version = %d, want %d", payload.FormatVersion, bidCodecVectorFormatVersion)
	}
	for _, anchor := range bidCodecVectorAnchors {
		found := false
		for _, v := range payload.Vectors {
			if v.Type != anchor.Type || v.Hex != anchor.Hex || v.HexHi != anchor.HexHi {
				continue
			}
			found = true
			if v.Sign != anchor.Sign || v.Coefficient != anchor.Coefficient ||
				v.Exponent != anchor.Exponent || v.Kind != anchor.Kind ||
				v.Payload != anchor.Payload || v.DecimalString != anchor.DecimalString ||
				v.Canonical != anchor.Canonical || v.EncodedHex != anchor.EncodedHex ||
				v.EncodedHi != anchor.EncodedHi {
				t.Errorf("anchor %s %s%s diverged from the checked-in vectors.json record: anchor=%+v record=%+v",
					anchor.Type, anchor.HexHi, anchor.Hex, anchor, v)
			}
			break
		}
		if !found {
			t.Errorf("anchor %s %s%s is missing from the checked-in vectors.json", anchor.Type, anchor.HexHi, anchor.Hex)
		}
	}
}
