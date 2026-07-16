package testgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadGeneratedRejectsUnknownFieldsAndTrailingJSON(t *testing.T) {
	manifest := Manifest{Output: "generated/testspec/spec_index.json"}
	files, err := EncodeSpecFiles(manifest, SharedSpec{
		ReadCases: []GeneratedReadCase{{
			Suite:    "strict_readtest",
			Format:   "decimal64",
			Header:   "readtest.h",
			Source:   "readtest.in",
			ID:       "strict_readtest_001",
			Line:     1,
			Function: "bid64_add",
			Kind:     "binary_op",
			Operands: []string{"1", "2"},
			Expected: "3",
			Status:   "00",
			Rounding: 0,
		}},
		FFICases: []GeneratedFFICase{{
			Suite:       "strict_ffi",
			ID:          "strict_ffi_001",
			Format:      "decimal64",
			ResultBits:  64,
			OperandBits: []int{128},
			Operation:   "sqrt",
			Function:    "bid64q_sqrt",
			LinkName:    "__bid64q_sqrt",
			Declaration: "BID_UINT64 bid64q_sqrt(...) ",
			Source:      "generated/json/intel_dfp_symbols.json",
			Rounding:    0,
			Operands:    []string{"30400000000000000000000000000002"},
		}},
	})
	if err != nil {
		t.Fatalf("EncodeSpecFiles: %v", err)
	}

	indexRel := filepath.ToSlash(manifest.Output)
	readtestRel := filepath.ToSlash(filepath.Join(filepath.Dir(manifest.Output), "readtest", "strict_readtest.json"))
	ffiRel := filepath.ToSlash(filepath.Join(filepath.Dir(manifest.Output), "ffi", "bid64q_sqrt.json"))
	for _, tc := range []struct {
		name    string
		target  string
		mutate  func([]byte) []byte
		wantErr string
	}{
		{name: "index unknown field", target: indexRel, mutate: addUnknownJSONField, wantErr: "unknown field"},
		{name: "index trailing value", target: indexRel, mutate: appendTrailingJSONValue, wantErr: "trailing JSON"},
		{name: "readtest shard unknown field", target: readtestRel, mutate: addUnknownJSONField, wantErr: "unknown field"},
		{name: "readtest shard trailing value", target: readtestRel, mutate: appendTrailingJSONValue, wantErr: "trailing JSON"},
		{name: "ffi shard unknown field", target: ffiRel, mutate: addUnknownJSONField, wantErr: "unknown field"},
		{name: "ffi shard trailing value", target: ffiRel, mutate: appendTrailingJSONValue, wantErr: "trailing JSON"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			for relative, data := range files {
				if relative == tc.target {
					data = tc.mutate(data)
				}
				full := filepath.Join(root, filepath.FromSlash(relative))
				if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
					t.Fatalf("mkdir %s: %v", full, err)
				}
				if err := os.WriteFile(full, data, 0o644); err != nil {
					t.Fatalf("write %s: %v", full, err)
				}
			}
			_, err := LoadGenerated(filepath.Join(root, filepath.FromSlash(indexRel)))
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("LoadGenerated error = %v, want rejection containing %q", err, tc.wantErr)
			}
		})
	}
}

func addUnknownJSONField(data []byte) []byte {
	return []byte(strings.Replace(string(data), "{", "{\n  \"unexpected_field\": true,", 1))
}

func appendTrailingJSONValue(data []byte) []byte {
	return append(append([]byte(nil), data...), []byte("{}\n")...)
}
