package testgen

import (
	"fmt"
	"os"
	"path/filepath"
)

func GenerateTier1Outputs() (map[string][]byte, error) {
	files, err := GenerateTier1PathsOutputs()
	if err != nil {
		return nil, err
	}
	java, err := GenerateTier1BigDecimalOutputs()
	if err != nil {
		return nil, err
	}
	for path, data := range java {
		if _, exists := files[path]; exists {
			return nil, fmt.Errorf("duplicate Tier1 generated output %s", path)
		}
		files[path] = data
	}
	return files, nil
}

func WriteTier1Outputs(repoRoot string) error {
	files, err := GenerateTier1Outputs()
	if err != nil {
		return err
	}
	for path, data := range files {
		full := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(full, data, 0644); err != nil {
			return err
		}
	}
	return nil
}
