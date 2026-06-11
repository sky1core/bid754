package cgen

import (
	"encoding/json"
	"fmt"
	"os"
)

type Manifest struct {
	GoPackage  string      `json:"go_package"`
	GoOutput   string      `json:"go_output"`
	RustOutput string      `json:"rust_output"`
	Tables     []TableSpec `json:"tables"`
}

type TableSpec struct {
	Name     string `json:"name"`
	Source   string `json:"source"`
	GoName   string `json:"go_name"`
	RustName string `json:"rust_name"`
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
	if len(manifest.Tables) == 0 {
		return manifest, fmt.Errorf("manifest %q: tables must not be empty", path)
	}

	for i, table := range manifest.Tables {
		if table.Name == "" || table.Source == "" || table.GoName == "" || table.RustName == "" {
			return manifest, fmt.Errorf("manifest %q: table %d is incomplete", path, i)
		}
	}

	return manifest, nil
}
