package csymbols

import (
	"encoding/json"
	"fmt"
	"os"
)

type Manifest struct {
	Library string       `json:"library"`
	Output  string       `json:"output"`
	Headers []HeaderSpec `json:"headers"`
}

type HeaderSpec struct {
	Path           string `json:"path"`
	ExtractSymbols bool   `json:"extract_symbols,omitempty"`
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
	if manifest.Library == "" {
		return manifest, fmt.Errorf("manifest %q: library is required", path)
	}
	if manifest.Output == "" {
		return manifest, fmt.Errorf("manifest %q: output is required", path)
	}
	if len(manifest.Headers) == 0 {
		return manifest, fmt.Errorf("manifest %q: headers must not be empty", path)
	}
	for i, header := range manifest.Headers {
		if header.Path == "" {
			return manifest, fmt.Errorf("manifest %q: header %d path is required", path, i)
		}
	}

	return manifest, nil
}
