package cgen

import (
	"encoding/json"
	"fmt"
	"os"
)

type Manifest struct {
	GoPackage   string      `json:"go_package"`
	GoOutput    string      `json:"go_output"`
	RustOutput  string      `json:"rust_output"`
	BidgoOutput string      `json:"bidgo_output"`
	BidgoSource string      `json:"bidgo_source"`
	Tables      []TableSpec `json:"tables"`
}

type TableSpec struct {
	Name     string `json:"name"`
	Source   string `json:"source"`
	GoName   string `json:"go_name"`
	RustName string `json:"rust_name"`
	// ExpectedShape pins the fully resolved dimensions of the parsed table
	// (outermost dimension first; fixed-word tuple arities such as BID_UINT128
	// are folded into the element type, matching Table.Dims). It is required in
	// the manifest so that a truncated or partially parsed initializer cannot
	// silently produce an empty or short table: ParseTableFile fails when the
	// parsed shape does not match this pin.
	ExpectedShape []int `json:"expected_shape"`
}

func LoadManifest(path string) (Manifest, error) {
	var manifest Manifest

	data, err := os.ReadFile(path)
	if err != nil {
		return manifest, fmt.Errorf("read manifest %q: %w", path, err)
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return manifest, fmt.Errorf("parse manifest %q: %w", path, err)
	}
	if manifest.GoPackage == "" {
		return manifest, fmt.Errorf("manifest %q: go_package is required", path)
	}
	if manifest.GoOutput == "" {
		return manifest, fmt.Errorf("manifest %q: go_output is required", path)
	}
	if manifest.RustOutput == "" {
		return manifest, fmt.Errorf("manifest %q: rust_output is required", path)
	}
	if manifest.BidgoOutput == "" {
		return manifest, fmt.Errorf("manifest %q: bidgo_output is required", path)
	}
	if manifest.BidgoSource == "" {
		return manifest, fmt.Errorf("manifest %q: bidgo_source is required", path)
	}
	if len(manifest.Tables) == 0 {
		return manifest, fmt.Errorf("manifest %q: tables must not be empty", path)
	}

	for i, table := range manifest.Tables {
		if table.Name == "" || table.Source == "" || table.GoName == "" || table.RustName == "" {
			return manifest, fmt.Errorf("manifest %q: table %d is incomplete", path, i)
		}
		if len(table.ExpectedShape) == 0 {
			return manifest, fmt.Errorf("manifest %q: table %q is missing expected_shape; pin the resolved dimensions so a truncated parse cannot pass silently", path, table.Name)
		}
		for d, n := range table.ExpectedShape {
			if n <= 0 {
				return manifest, fmt.Errorf("manifest %q: table %q expected_shape[%d] = %d must be > 0", path, table.Name, d, n)
			}
		}
	}

	return manifest, nil
}
